/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18014f130 */

int * FUN_18014f130(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  undefined8 *puVar3;
  longlong *plVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  code *pcVar8;
  code *pcVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  int *piVar13;
  undefined8 uVar14;
  longlong *local_res18;
  longlong *local_58;
  longlong *local_50;
  longlong *local_40;

  plVar11 = DAT_1803b6b40;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  local_40 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_58 = DAT_180391d20;
  if (DAT_1803b6b40 == (longlong *)0x0) {
LAB_18014f1aa:
    DAT_1803b6b40 = (longlong *)FUN_18036cd10(param_1,DAT_1803b72c0,DAT_1803ca758,0x20);
  }
  else if ((1 < *DAT_1803b6b40) || (DAT_1803b6b40[2] != 0)) {
    if ((DAT_1803b6b40 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b40 && (*DAT_1803b6b40 = *DAT_1803b6b40 + -1, *plVar11 == 0)))) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    goto LAB_18014f1aa;
  }
  plVar4 = DAT_1803b6b40;
  FUN_180001d00(param_1,DAT_1803b6b40);
  lVar6 = FUN_18014c500(param_1);
  plVar11 = (longlong *)0x0;
  if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
    FUN_180373340(param_1,&local_58,DAT_1803b6c30);
    uVar14 = 0x58;
    plVar12 = local_58;
  }
  else {
    plVar7 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.short_fiction_parsers#22="^##\\s*(?:第\\s*" */
    piVar13 = DAT_1803b6ce0;
    if (plVar7 == (longlong *)0x0) {
      plVar12 = *(longlong **)(param_1 + 0x60);
      uVar14 = 0x58;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar6 = *(longlong *)(param_1 + 0x10);
      plVar11 = *(longlong **)(lVar6 + 0x41da8);
      if (plVar11 == (longlong *)0x0) {
        plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
      }
      else {
        *(longlong *)(lVar6 + 0x41da8) = plVar11[3];
        *(int *)(lVar6 + 0x41e40) = *(int *)(lVar6 + 0x41e40) + -1;
        *plVar11 = 1;
      }
      plVar11[3] = 0;
      plVar10 = plVar11 + -2;
      plVar11[4] = 0;
      plVar11[5] = 0;
      lVar6 = _PyThreadState_GetCurrent();
      lVar6 = *(longlong *)(*(longlong *)(lVar6 + 0x10) + 0xd0);
      puVar3 = *(undefined8 **)(lVar6 + 8);
      *puVar3 = plVar10;
      plVar11[-1] = plVar11[-1] & 3;
      plVar11[-1] = plVar11[-1] | (ulonglong)puVar3;
      *plVar10 = lVar6;
      *(longlong **)(lVar6 + 8) = plVar10;
      if (*piVar13 + 1 != 0) {
        *piVar13 = *piVar13 + 1;
      }
      plVar11[3] = (longlong)piVar13;
/* novels.pipeline.short_fiction_parsers#4="" */
      lVar6 = FUN_180375520(param_1,plVar2,DAT_1803b6c50);
      if (lVar6 == 0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        uVar14 = 0x58;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        plVar11 = (longlong *)0x0;
      }
      else {
        plVar11[4] = lVar6;
/* novels.pipeline.short_fiction_parsers#23="\\s*章\\s*)?(.+)$" */
        piVar13 = DAT_1803b6ce8;
/* novels.pipeline.short_fiction_parsers#23="\\s*章\\s*)?(.+)$" */
        if (*DAT_1803b6ce8 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#23="\\s*章\\s*)?(.+)$" */
          *DAT_1803b6ce8 = *DAT_1803b6ce8 + 1;
        }
        plVar11[5] = (longlong)piVar13;
/* novels.pipeline.short_fiction_parsers#4="" */
        plVar10 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar11);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (plVar10 == (longlong *)0x0) {
          plVar12 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
LAB_18014f3c5:
          uVar14 = 0x58;
          plVar11 = local_res18;
        }
        else {
          lVar6 = FUN_18014c500(param_1);
          if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
            FUN_180373340(param_1,&local_58,DAT_1803b6c30);
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            uVar14 = 0x58;
            plVar11 = (longlong *)0x0;
            plVar12 = local_58;
          }
          else {
            plVar11 = (longlong *)FUN_18037f0c0();
            if (plVar11 == (longlong *)0x0) {
              plVar12 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              if (((int)*plVar10 < 0) || (*plVar10 = *plVar10 + -1, *plVar10 != 0))
              goto LAB_18014f3c5;
              (**(code **)(plVar10[1] + 0x30))(plVar10);
              plVar11 = (longlong *)0x0;
              uVar14 = 0x58;
            }
            else {
              uVar14 = 0x58;
              *(undefined4 *)(plVar4 + 5) = 0x58;
              local_58 = plVar10;
              local_50 = plVar11;
              plVar12 = (longlong *)FUN_180377750(param_1,plVar7,&local_58);
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
              if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                (**(code **)(plVar11[1] + 0x30))(plVar11);
              }
              if (plVar12 != (longlong *)0x0) {
                uVar14 = 0x59;
                *(undefined4 *)(plVar4 + 5) = 0x59;
                local_40 = plVar12;
/* novels.pipeline.short_fiction_parsers#14="search" */
                plVar11 = (longlong *)FUN_18037e6c0(param_1,plVar12,DAT_1803b6ca0,plVar1);
                if (plVar11 != (longlong *)0x0) {
                  iVar5 = FUN_180001510(plVar11);
/* novels.pipeline.short_fiction_parsers#4="" */
                  piVar13 = DAT_1803b6c50;
                  local_res18 = plVar11;
                  if (iVar5 == -1) {
                    uVar14 = 0x5a;
                  }
                  else {
                    if (iVar5 == 0) {
/* novels.pipeline.short_fiction_parsers#4="" */
                      if (*DAT_1803b6c50 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#4="" */
                        *DAT_1803b6c50 = *DAT_1803b6c50 + 1;
                      }
LAB_18014f649:
                      FUN_180001d80(param_1);
                      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                        (**(code **)(plVar12[1] + 0x30))(plVar12);
                      }
                      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                      }
                      if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                        (**(code **)(plVar1[1] + 0x30))(plVar1);
                      }
                      if ((int)*plVar2 < 0) {
                        return piVar13;
                      }
                      *plVar2 = *plVar2 + -1;
                      if (*plVar2 != 0) {
                        return piVar13;
                      }
                      (**(code **)(plVar2[1] + 0x30))(plVar2);
                      return piVar13;
                    }
                    uVar14 = 0x5a;
                    *(undefined4 *)(plVar4 + 5) = 0x5a;
                    plVar7 = (longlong *)
/* novels.pipeline.short_fiction_parsers#20="group" */
                             FUN_18037e6c0(param_1,plVar11,DAT_1803b6cd0,
/* novels.pipeline.short_fiction_parsers#21={"type": "tuple", "items": [{"tag": "l", "fileRange": [3998986, 3998988], "value": 1}]} */
                                           *(undefined8 *)(DAT_1803b6cd8 + 0x18));
                    if (plVar7 != (longlong *)0x0) {
                      plVar10 = (longlong *)FUN_18037f0c0();
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      if (plVar10 == (longlong *)0x0) {
                        plVar12 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        goto LAB_18014f32a;
                      }
                      *(undefined4 *)(plVar4 + 5) = 0x5a;
                      piVar13 = (int *)FUN_180376210(param_1,plVar10);
                      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                      }
                      if (piVar13 != (int *)0x0) goto LAB_18014f649;
                    }
                  }
                }
              }
              plVar12 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar11 = local_res18;
            }
          }
        }
      }
    }
  }
LAB_18014f32a:
  piVar13 = (int *)plVar12[5];
  if (piVar13 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar4,uVar14);
    plVar7 = (longlong *)plVar12[5];
    pcVar9 = pcVar8;
    if (pcVar8 == _Py_NoneStruct_exref) {
      pcVar9 = (code *)0x0;
    }
    else {
LAB_18014f707:
      pcVar8 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar13 + 6) == plVar4) goto LAB_18014f750;
    pcVar8 = (code *)FUN_18038e2a0(plVar4,uVar14);
    *(int **)(pcVar8 + 0x10) = piVar13;
    if (*piVar13 + 1 != 0) {
      *piVar13 = *piVar13 + 1;
    }
    plVar7 = (longlong *)plVar12[5];
    pcVar9 = pcVar8;
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_18014f707;
    pcVar9 = (code *)0x0;
  }
  plVar12[5] = (longlong)pcVar9;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
LAB_18014f750:
  plVar10 = local_40;
  FUN_18036d0a0(plVar4,&DAT_180391c1c,plVar1,plVar2,local_40,plVar11);
  plVar7 = DAT_1803b6b40;
  if (plVar4 == DAT_1803b6b40) {
    if ((-1 < (int)*DAT_1803b6b40) && (*DAT_1803b6b40 = *DAT_1803b6b40 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b6b40 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (int *)0x0;
}
