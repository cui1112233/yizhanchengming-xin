/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180152b30 */

longlong * FUN_180152b30(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  undefined8 *puVar3;
  int *piVar4;
  longlong lVar5;
  longlong *plVar6;
  int iVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  code *pcVar12;
  code *pcVar13;
  longlong *plVar14;
  undefined8 uVar15;
  longlong local_58;
  longlong *local_50;
  longlong *local_48;

  plVar6 = DAT_1803b6b10;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  local_58 = DAT_180391d20;
  if (DAT_1803b6b10 == (longlong *)0x0) {
LAB_180152b94:
    DAT_1803b6b10 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b80,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6b10) || (DAT_1803b6b10[2] != 0)) {
    if ((DAT_1803b6b10 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b10 && (*DAT_1803b6b10 = *DAT_1803b6b10 + -1, *plVar6 == 0)))) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    goto LAB_180152b94;
  }
  plVar6 = DAT_1803b6b10;
  FUN_180001d00(param_1,DAT_1803b6b10);
  plVar8 = (longlong *)FUN_18037f0c0();
  uVar15 = 0x8c;
  if (plVar8 == (longlong *)0x0) {
    lVar11 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar9 = (longlong *)0x0;
  }
  else {
    *(undefined4 *)(plVar6 + 5) = 0x8c;
    plVar9 = (longlong *)FUN_180376210(param_1,plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (plVar9 == (longlong *)0x0) {
      lVar11 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar9 = (longlong *)0x0;
    }
    else {
      iVar7 = FUN_180001510(plVar9);
/* novels.pipeline.short_fiction_parsers#40="第" */
      piVar4 = DAT_1803b6d70;
      if (iVar7 == -1) {
        lVar11 = *(longlong *)(param_1 + 0x60);
        uVar15 = 0x8d;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else if (iVar7 == 0) {
        lVar11 = *(longlong *)(param_1 + 0x10);
        plVar10 = *(longlong **)(lVar11 + 0x41da8);
        if (plVar10 == (longlong *)0x0) {
          plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
        }
        else {
          *(longlong *)(lVar11 + 0x41da8) = plVar10[3];
          *(int *)(lVar11 + 0x41e40) = *(int *)(lVar11 + 0x41e40) + -1;
          *plVar10 = 1;
        }
        plVar8 = plVar10 + -2;
        plVar10[3] = 0;
        plVar10[4] = 0;
        plVar10[5] = 0;
        lVar11 = _PyThreadState_GetCurrent();
        lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
        puVar3 = *(undefined8 **)(lVar11 + 8);
        *puVar3 = plVar8;
        plVar10[-1] = plVar10[-1] & 3;
        plVar10[-1] = plVar10[-1] | (ulonglong)puVar3;
        *plVar8 = lVar11;
        *(longlong **)(lVar11 + 8) = plVar8;
        if (*piVar4 + 1 != 0) {
          *piVar4 = *piVar4 + 1;
        }
        plVar10[3] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
        lVar11 = FUN_180375520(param_1,plVar1,DAT_1803b6c50);
        if (lVar11 == 0) {
          lVar11 = *(longlong *)(param_1 + 0x60);
          uVar15 = 0x8e;
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
        }
        else {
          plVar10[4] = lVar11;
/* novels.pipeline.short_fiction_parsers#41="章" */
          piVar4 = DAT_1803b6d78;
/* novels.pipeline.short_fiction_parsers#41="章" */
          if (*DAT_1803b6d78 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#41="章" */
            *DAT_1803b6d78 = *DAT_1803b6d78 + 1;
          }
          plVar10[5] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
          plVar8 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar10);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (plVar8 != (longlong *)0x0) goto LAB_180153373;
          lVar11 = *(longlong *)(param_1 + 0x60);
          uVar15 = 0x8e;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
      }
      else {
        lVar11 = FUN_18014c500(param_1);
        if (lVar11 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
          FUN_180373340(param_1,&local_58,DAT_1803b6c30);
          uVar15 = 0x8f;
          lVar11 = local_58;
        }
        else {
          plVar8 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.short_fiction_parsers#37="^第\\s*" */
          piVar4 = DAT_1803b6d58;
          if (plVar8 == (longlong *)0x0) {
            lVar11 = *(longlong *)(param_1 + 0x60);
            uVar15 = 0x8f;
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            lVar11 = *(longlong *)(param_1 + 0x10);
            plVar10 = *(longlong **)(lVar11 + 0x41da8);
            if (plVar10 == (longlong *)0x0) {
              plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
            }
            else {
              *(longlong *)(lVar11 + 0x41da8) = plVar10[3];
              *(int *)(lVar11 + 0x41e40) = *(int *)(lVar11 + 0x41e40) + -1;
              *plVar10 = 1;
            }
            plVar14 = plVar10 + -2;
            plVar10[3] = 0;
            plVar10[4] = 0;
            plVar10[5] = 0;
            lVar11 = _PyThreadState_GetCurrent();
            lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
            puVar3 = *(undefined8 **)(lVar11 + 8);
            *puVar3 = plVar14;
            plVar10[-1] = plVar10[-1] & 3;
            plVar10[-1] = plVar10[-1] | (ulonglong)puVar3;
            *plVar14 = lVar11;
            *(longlong **)(lVar11 + 8) = plVar14;
            if (*piVar4 + 1 != 0) {
              *piVar4 = *piVar4 + 1;
            }
            plVar10[3] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
            lVar11 = FUN_180375520(param_1,plVar1,DAT_1803b6c50);
            if (lVar11 == 0) {
              lVar11 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
LAB_180152ec9:
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
            }
            else {
              plVar10[4] = lVar11;
/* novels.pipeline.short_fiction_parsers#43="\\s*章" */
              piVar4 = DAT_1803b6d88;
/* novels.pipeline.short_fiction_parsers#43="\\s*章" */
              if (*DAT_1803b6d88 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#43="\\s*章" */
                *DAT_1803b6d88 = *DAT_1803b6d88 + 1;
              }
              plVar10[5] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
              plVar14 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar10);
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
              if (plVar14 == (longlong *)0x0) {
                lVar11 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
              }
              else {
                *(undefined4 *)(plVar6 + 5) = 0x8f;
                local_50 = plVar14;
                local_48 = plVar9;
                plVar10 = (longlong *)FUN_180377750(param_1,plVar8,&local_50);
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                }
                if (plVar10 != (longlong *)0x0) {
                  iVar7 = FUN_180001510(plVar10);
                  if (iVar7 != -1) {
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
/* novels.pipeline.short_fiction_parsers#40="第" */
                    piVar4 = DAT_1803b6d70;
                    if (iVar7 != 0) {
                      iVar7 = (int)*plVar9 + 1;
                      plVar8 = plVar9;
                      if (iVar7 != 0) {
                        *(int *)plVar9 = iVar7;
                      }
LAB_180153373:
                      FUN_180001d80(param_1);
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                        (**(code **)(plVar1[1] + 0x30))(plVar1);
                      }
                      if ((int)*plVar2 < 0) {
                        return plVar8;
                      }
                      *plVar2 = *plVar2 + -1;
                      if (*plVar2 != 0) {
                        return plVar8;
                      }
                      (**(code **)(plVar2[1] + 0x30))(plVar2);
                      return plVar8;
                    }
                    lVar11 = *(longlong *)(param_1 + 0x10);
                    plVar10 = *(longlong **)(lVar11 + 0x41db0);
                    if (plVar10 == (longlong *)0x0) {
                      plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,4);
                    }
                    else {
                      *(longlong *)(lVar11 + 0x41db0) = plVar10[3];
                      *(int *)(lVar11 + 0x41e44) = *(int *)(lVar11 + 0x41e44) + -1;
                      *plVar10 = 1;
                    }
                    plVar8 = plVar10 + -2;
                    plVar10[3] = 0;
                    plVar10[4] = 0;
                    plVar10[5] = 0;
                    plVar10[6] = 0;
                    lVar11 = _PyThreadState_GetCurrent();
                    lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
                    puVar3 = *(undefined8 **)(lVar11 + 8);
                    *puVar3 = plVar8;
                    plVar10[-1] = plVar10[-1] & 3;
                    plVar10[-1] = plVar10[-1] | (ulonglong)puVar3;
                    *plVar8 = lVar11;
                    *(longlong **)(lVar11 + 8) = plVar8;
                    if (*piVar4 + 1 != 0) {
                      *piVar4 = *piVar4 + 1;
                    }
                    plVar10[3] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
                    lVar11 = FUN_180375520(param_1,plVar1,DAT_1803b6c50);
                    if (lVar11 != 0) {
                      plVar10[4] = lVar11;
/* novels.pipeline.short_fiction_parsers#44="章 " */
                      piVar4 = DAT_1803b6d90;
/* novels.pipeline.short_fiction_parsers#44="章 " */
                      if (*DAT_1803b6d90 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#44="章 " */
                        *DAT_1803b6d90 = *DAT_1803b6d90 + 1;
                      }
                      plVar10[5] = (longlong)piVar4;
/* novels.pipeline.short_fiction_parsers#4="" */
                      lVar11 = FUN_180375520(param_1,plVar9,DAT_1803b6c50);
                      if (lVar11 != 0) {
                        plVar10[6] = lVar11;
/* novels.pipeline.short_fiction_parsers#4="" */
                        plVar8 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar10);
                        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                        if (plVar8 != (longlong *)0x0) goto LAB_180153373;
                        lVar11 = *(longlong *)(param_1 + 0x60);
                        uVar15 = 0x91;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        goto LAB_180152eea;
                      }
                    }
                    uVar15 = 0x91;
                    lVar11 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                    goto LAB_180152eea;
                  }
                  lVar11 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_180152ec9;
                }
                lVar11 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
            }
            uVar15 = 0x8f;
          }
        }
      }
    }
  }
LAB_180152eea:
  piVar4 = *(int **)(lVar11 + 0x28);
  if (piVar4 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar6,uVar15);
    plVar8 = *(longlong **)(lVar11 + 0x28);
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_180153208:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar6) goto LAB_180153251;
    pcVar12 = (code *)FUN_18038e2a0(plVar6,uVar15);
    *(int **)(pcVar12 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar8 = *(longlong **)(lVar11 + 0x28);
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_180153208;
    pcVar13 = (code *)0x0;
  }
  *(code **)(lVar11 + 0x28) = pcVar13;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_180153251:
  FUN_18036d0a0(plVar6,&DAT_1803948e8,plVar1,plVar2,plVar9);
  plVar8 = DAT_1803b6b10;
  if (plVar6 == DAT_1803b6b10) {
    if ((-1 < (int)*DAT_1803b6b10) && (*DAT_1803b6b10 = *DAT_1803b6b10 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803b6b10 = (longlong *)0x0;
  }
  lVar5 = **(longlong **)(param_1 + 0x38);
  plVar6 = *(longlong **)(lVar5 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar5 + 8);
  *(undefined4 *)(plVar6 + 8) = 0xffffffff;
  plVar8 = (longlong *)plVar6[2];
  if (((plVar8 != (longlong *)0x0) && (plVar6[2] = 0, -1 < (int)*plVar8)) &&
     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  plVar6[10] = 0;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar11;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
