/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180157690 */

code * FUN_180157690(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  code *pcVar2;
  int *piVar3;
  longlong *plVar4;
  uint uVar5;
  int iVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong lVar9;
  longlong *plVar10;
  longlong *plVar11;
  undefined8 *puVar12;
  longlong *plVar13;
  code *pcVar14;
  code *pcVar15;
  undefined8 uVar16;
  longlong *local_res18;
  longlong local_res20;

  plVar4 = DAT_1803b6bb0;
  plVar11 = (longlong *)0x0;
  plVar1 = (longlong *)*param_3;
  pcVar2 = (code *)param_3[1];
  local_res18 = (longlong *)0x0;
  local_res20 = DAT_180391d20;
  if (DAT_1803b6bb0 == (longlong *)0x0) {
LAB_1801576fc:
    DAT_1803b6bb0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6a38,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6bb0) || (DAT_1803b6bb0[2] != 0)) {
    if ((DAT_1803b6bb0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6bb0 && (*DAT_1803b6bb0 = *DAT_1803b6bb0 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1801576fc;
  }
  plVar4 = DAT_1803b6bb0;
  FUN_180001d00(param_1,DAT_1803b6bb0);
  uVar5 = (uint)(pcVar2 != _Py_NoneStruct_exref);
  if (pcVar2 != _Py_NoneStruct_exref) {
    plVar7 = (longlong *)FUN_18037f0c0();
    if (plVar7 == (longlong *)0x0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      uVar16 = 0x103;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar8 = FUN_180373860(plVar7);
      plVar10 = (longlong *)(param_1 + 0x60);
      if ((-1 < lVar8) || (*plVar10 == 0)) {
        plVar11 = (longlong *)PyLong_FromSsize_t(lVar8);
      }
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (plVar11 == (longlong *)0x0) {
        lVar8 = *plVar10;
        *plVar10 = 0;
        uVar16 = 0x103;
        plVar11 = local_res18;
      }
      else {
        uVar5 = FUN_180386680(plVar11,pcVar2);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (uVar5 != 0xffffffff) goto LAB_18015781a;
        lVar8 = *(longlong *)(param_1 + 0x60);
        uVar16 = 0x103;
        *(longlong *)(param_1 + 0x60) = 0;
        plVar11 = local_res18;
      }
    }
    goto LAB_180157f3a;
  }
LAB_18015781a:
  plVar11 = (longlong *)0x0;
  if (uVar5 == 1) {
    lVar8 = FUN_18014bc50(param_1);
/* novels.pipeline.short_fiction_parsers#83="短篇草稿不完整：期望 " */
    piVar3 = DAT_1803b6ec8;
    if (lVar8 == 0) {
/* novels.pipeline.short_fiction_parsers#82="ShortFictionDraftValidationError" */
      FUN_180373340(param_1,&local_res20,DAT_1803b6ec0);
      uVar16 = 0x104;
      lVar8 = local_res20;
      plVar11 = local_res18;
      goto LAB_180157f3a;
    }
    lVar9 = *(longlong *)(param_1 + 0x10);
    plVar7 = *(longlong **)(lVar9 + 0x41db8);
    if (plVar7 == (longlong *)0x0) {
      plVar7 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,5);
    }
    else {
      *(longlong *)(lVar9 + 0x41db8) = plVar7[3];
      *(int *)(lVar9 + 0x41e48) = *(int *)(lVar9 + 0x41e48) + -1;
      *plVar7 = 1;
    }
    plVar7[3] = 0;
    plVar10 = plVar7 + -2;
    plVar7[4] = 0;
    plVar7[5] = 0;
    plVar7[6] = 0;
    plVar7[7] = 0;
    lVar9 = _PyThreadState_GetCurrent();
    lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
    puVar12 = *(undefined8 **)(lVar9 + 8);
    *puVar12 = plVar10;
    plVar7[-1] = plVar7[-1] & 3;
    plVar7[-1] = plVar7[-1] | (ulonglong)puVar12;
    *plVar10 = lVar9;
    *(longlong **)(lVar9 + 8) = plVar10;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7[3] = (longlong)piVar3;
/* novels.pipeline.short_fiction_parsers#4="" */
    lVar9 = FUN_180375520(param_1,pcVar2,DAT_1803b6c50);
    if (lVar9 == 0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1801579eb:
      uVar16 = 0x105;
      plVar11 = local_res18;
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
    }
    else {
      plVar7[4] = lVar9;
/* novels.pipeline.short_fiction_parsers#84=" 章，实际 " */
      piVar3 = DAT_1803b6ed0;
/* novels.pipeline.short_fiction_parsers#84=" 章，实际 " */
      if (*DAT_1803b6ed0 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#84=" 章，实际 " */
        *DAT_1803b6ed0 = *DAT_1803b6ed0 + 1;
      }
      plVar7[5] = (longlong)piVar3;
/* novels.pipeline.short_fiction_parsers#63="chapters" */
      plVar10 = (longlong *)FUN_18037f0c0(piVar3,plVar1,DAT_1803b6e28);
      if (plVar10 == (longlong *)0x0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1801579eb;
      }
      lVar9 = FUN_180373860(plVar10);
      plVar13 = (longlong *)(param_1 + 0x60);
      if ((-1 < lVar9) || (*plVar13 == 0)) {
        plVar11 = (longlong *)PyLong_FromSsize_t(lVar9);
      }
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (plVar11 == (longlong *)0x0) {
        lVar8 = *plVar13;
        *plVar13 = 0;
        goto LAB_1801579eb;
      }
/* novels.pipeline.short_fiction_parsers#4="" */
      lVar9 = FUN_180375520(param_1,plVar11,DAT_1803b6c50);
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (lVar9 == 0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(longlong *)(param_1 + 0x60) = 0;
        goto LAB_1801579eb;
      }
      plVar7[6] = lVar9;
/* novels.pipeline.short_fiction_parsers#85=" 章。" */
      piVar3 = DAT_1803b6ed8;
/* novels.pipeline.short_fiction_parsers#85=" 章。" */
      if (*DAT_1803b6ed8 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#85=" 章。" */
        *DAT_1803b6ed8 = *DAT_1803b6ed8 + 1;
      }
      plVar7[7] = (longlong)piVar3;
/* novels.pipeline.short_fiction_parsers#4="" */
      plVar11 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar7);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (plVar11 == (longlong *)0x0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(longlong *)(param_1 + 0x60) = 0;
        uVar16 = 0x105;
        plVar11 = local_res18;
        goto LAB_180157f3a;
      }
      uVar16 = 0x104;
      *(undefined4 *)(plVar4 + 5) = 0x104;
      plVar7 = (longlong *)FUN_180376770(param_1,lVar8,plVar11);
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (plVar7 == (longlong *)0x0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(longlong *)(param_1 + 0x60) = 0;
        plVar11 = (longlong *)0x0;
        goto LAB_180157f3a;
      }
      lVar8 = FUN_180373490(param_1,plVar7);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (lVar8 == 0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(longlong *)(param_1 + 0x60) = 0;
        plVar11 = local_res18;
        goto LAB_180157f3a;
      }
LAB_180157f27:
      FUN_18000d1d0(param_1,lVar8);
      plVar11 = local_res18;
    }
  }
  else {
    lVar8 = *(longlong *)(DAT_1803ca760 + 0x20);
    if (*(char *)(lVar8 + 10) == '\0') {
/* novels.pipeline.short_fiction_parsers#86="find_empty_short_fiction_chapters" */
      puVar12 = (undefined8 *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6ee0);
      if (puVar12 != (undefined8 *)0x0) {
        plVar11 = (longlong *)*puVar12;
      }
LAB_180157be0:
      if (plVar11 == (longlong *)0x0) goto LAB_180157be5;
    }
    else {
      iVar6 = *(int *)(lVar8 + 0xc);
      if (iVar6 == 0) {
        iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
        *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar6 + 1;
        *(int *)(lVar8 + 0xc) = iVar6;
      }
      if (iVar6 != DAT_1803a98e8) {
        DAT_1803a98e8 = iVar6;
/* novels.pipeline.short_fiction_parsers#86="find_empty_short_fiction_chapters" */
        DAT_1803b72d0 = FUN_1803707d0(lVar8,DAT_1803b6ee0,*(undefined8 *)(DAT_1803b6ee0 + 0x18));
      }
      if (-1 < DAT_1803b72d0) {
        lVar9 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
        plVar11 = *(longlong **)(lVar9 + 0x28 + DAT_1803b72d0 * 0x10);
        if (plVar11 != (longlong *)0x0) goto LAB_180157c2e;
/* novels.pipeline.short_fiction_parsers#86="find_empty_short_fiction_chapters" */
        DAT_1803b72d0 = FUN_1803707d0(lVar8,DAT_1803b6ee0,*(undefined8 *)(DAT_1803b6ee0 + 0x18));
        if (-1 < DAT_1803b72d0) {
          plVar11 = *(longlong **)(lVar9 + 0x28 + DAT_1803b72d0 * 0x10);
          goto LAB_180157be0;
        }
      }
LAB_180157be5:
/* novels.pipeline.short_fiction_parsers#86="find_empty_short_fiction_chapters" */
      puVar12 = (undefined8 *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6ee0);
      if ((puVar12 == (undefined8 *)0x0) ||
         (plVar11 = (longlong *)*puVar12, plVar11 == (longlong *)0x0)) {
/* novels.pipeline.short_fiction_parsers#86="find_empty_short_fiction_chapters" */
        FUN_180373340(param_1,&local_res20,DAT_1803b6ee0);
        uVar16 = 0x107;
        lVar8 = local_res20;
        plVar11 = local_res18;
        goto LAB_180157f3a;
      }
    }
LAB_180157c2e:
    uVar16 = 0x107;
    *(undefined4 *)(plVar4 + 5) = 0x107;
    plVar11 = (longlong *)FUN_180376770(param_1,plVar11,plVar1);
    if (plVar11 == (longlong *)0x0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar11 = local_res18;
    }
    else {
      iVar6 = FUN_180001510(plVar11);
      if (iVar6 == -1) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        uVar16 = 0x108;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        if (iVar6 == 0) {
          FUN_180001d80(param_1);
          pcVar14 = _Py_NoneStruct_exref;
          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
            *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
          }
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
            (**(code **)(plVar1[1] + 0x30))(plVar1);
          }
          if (*(int *)pcVar2 < 0) {
            return pcVar14;
          }
          *(longlong *)pcVar2 = *(longlong *)pcVar2 + -1;
          if (*(longlong *)pcVar2 != 0) {
            return pcVar14;
          }
          (**(code **)(*(longlong *)(pcVar2 + 8) + 0x30))(pcVar2);
          return pcVar14;
        }
        lVar8 = FUN_18014bc50();
/* novels.pipeline.short_fiction_parsers#87="短篇草稿不完整：以下章节为空：" */
        piVar3 = DAT_1803b6ee8;
        if (lVar8 == 0) {
/* novels.pipeline.short_fiction_parsers#82="ShortFictionDraftValidationError" */
          FUN_180373340(param_1,&local_res20,DAT_1803b6ec0);
          uVar16 = 0x109;
          lVar8 = local_res20;
        }
        else {
          lVar9 = *(longlong *)(param_1 + 0x10);
          plVar7 = *(longlong **)(lVar9 + 0x41da8);
          if (plVar7 == (longlong *)0x0) {
            plVar7 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
          }
          else {
            *(longlong *)(lVar9 + 0x41da8) = plVar7[3];
            *(int *)(lVar9 + 0x41e40) = *(int *)(lVar9 + 0x41e40) + -1;
            *plVar7 = 1;
          }
          plVar10 = plVar7 + -2;
          plVar7[3] = 0;
          plVar7[4] = 0;
          plVar7[5] = 0;
          lVar9 = _PyThreadState_GetCurrent();
          lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
          puVar12 = *(undefined8 **)(lVar9 + 8);
          *puVar12 = plVar10;
          plVar7[-1] = plVar7[-1] & 3;
          plVar7[-1] = plVar7[-1] | (ulonglong)puVar12;
          *plVar10 = lVar9;
          *(longlong **)(lVar9 + 8) = plVar10;
          if (*piVar3 + 1 != 0) {
            *piVar3 = *piVar3 + 1;
          }
          plVar7[3] = (longlong)piVar3;
/* novels.pipeline.short_fiction_parsers#88=", " */
          uVar16 = DAT_1803b6ef0;
          plVar10 = (longlong *)FUN_18000d280(puVar12,plVar11);
          if (plVar10 != (longlong *)0x0) {
            local_res20 = FUN_18036d750(plVar10);
            plVar13 = (longlong *)
/* novels.pipeline.short_fiction_parsers#90="<genexpr>" ; novels.pipeline.short_fiction_parsers#91="validate_short_fiction_draft_for_final.<locals>.<genexpr>" */
                      FUN_180367890(FUN_180158130,DAT_1803ca758,DAT_1803b6f00,DAT_1803b6f08,
                                    DAT_1803b6a28,&local_res20,1,0x448);
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            lVar9 = PyUnicode_Join(uVar16,plVar13);
            if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
              (**(code **)(plVar13[1] + 0x30))(plVar13);
            }
            if (lVar9 != 0) {
              plVar7[4] = lVar9;
/* novels.pipeline.short_fiction_parsers#89="。" */
              piVar3 = DAT_1803b6ef8;
/* novels.pipeline.short_fiction_parsers#89="。" */
              if (*DAT_1803b6ef8 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#89="。" */
                *DAT_1803b6ef8 = *DAT_1803b6ef8 + 1;
              }
              plVar7[5] = (longlong)piVar3;
/* novels.pipeline.short_fiction_parsers#4="" */
              plVar10 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar7);
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              if (plVar10 == (longlong *)0x0) {
                lVar8 = *(longlong *)(param_1 + 0x60);
                uVar16 = 0x10a;
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
                uVar16 = 0x109;
                *(undefined4 *)(plVar4 + 5) = 0x109;
                plVar7 = (longlong *)FUN_180376770(param_1,lVar8,plVar10);
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (plVar7 == (longlong *)0x0) {
                  lVar8 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  lVar8 = FUN_180373490(param_1,plVar7);
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  local_res18 = plVar11;
                  if (lVar8 != 0) goto LAB_180157f27;
                  lVar8 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
              }
              goto LAB_180157f3a;
            }
          }
          uVar16 = 0x10a;
          lVar8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
        }
      }
    }
  }
LAB_180157f3a:
  piVar3 = *(int **)(lVar8 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar4,uVar16);
    plVar7 = *(longlong **)(lVar8 + 0x28);
    pcVar15 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
LAB_180157f94:
      pcVar14 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_180157fdd;
    pcVar14 = (code *)FUN_18038e2a0(plVar4,uVar16);
    *(int **)(pcVar14 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7 = *(longlong **)(lVar8 + 0x28);
    pcVar15 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_180157f94;
    pcVar15 = (code *)0x0;
  }
  *(code **)(lVar8 + 0x28) = pcVar15;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_180157fdd:
  FUN_18036d0a0(plVar4,&DAT_1803948e8,plVar1,pcVar2,plVar11);
  plVar7 = DAT_1803b6bb0;
  if (plVar4 == DAT_1803b6bb0) {
    if ((-1 < (int)*DAT_1803b6bb0) && (*DAT_1803b6bb0 = *DAT_1803b6bb0 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b6bb0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < *(int *)pcVar2) &&
     (*(longlong *)pcVar2 = *(longlong *)pcVar2 + -1, *(longlong *)pcVar2 == 0)) {
    (**(code **)(*(longlong *)(pcVar2 + 8) + 0x30))(pcVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
