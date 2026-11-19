/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c37d0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1802c37d0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  longlong *plVar2;
  undefined8 uVar3;
  char cVar4;
  int iVar5;
  code *pcVar6;
  longlong lVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  longlong lVar12;
  longlong lVar13;
  code *pcVar14;
  code *pcVar15;
  code *pcVar16;
  undefined8 uVar17;
  bool bVar18;
  undefined1 auStack_148 [32];
  code *local_128;
  code *local_120;
  code *local_118;
  code *local_110;
  code *local_108;
  code *local_100;
  code *local_f0;
  longlong *local_e8;
  code *local_e0;
  code *local_d8;
  code *local_d0;
  longlong *local_c8;
  code *local_c0;
  code *local_b8;
  code *local_b0;
  code *local_a8;
  code *local_a0;
  longlong *local_98;
  longlong *local_90;
  code *local_88;
  code *local_80;
  undefined8 local_78;
  code *local_70;
  code *local_68;
  code *local_60;
  code *local_58;
  code *local_50;
  code *local_48;
  ulonglong local_40;
  
  plVar8 = DAT_1803c4b78;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_148;
  local_98 = (longlong *)param_3[1];
  pcVar11 = (code *)*param_3;
  pcVar6 = (code *)0x0;
  local_d0 = (code *)0x0;
  local_c0 = (code *)0x0;
  local_b0 = (code *)0x0;
  local_b8 = (code *)0x0;
  local_a8 = (code *)0x0;
  local_c8 = DAT_180391e98;
  local_e8 = DAT_1803c4b78;
  local_a0 = pcVar11;
  if (DAT_1803c4b78 == (longlong *)0x0) {
LAB_1802c3872:
    DAT_1803c4b78 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4be0,DAT_1803ca468,0x40);
  }
  else if ((1 < *DAT_1803c4b78) || (DAT_1803c4b78[2] != 0)) {
    if ((DAT_1803c4b78 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4b78 && (*DAT_1803c4b78 = *DAT_1803c4b78 + -1, *plVar8 == 0)))) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    goto LAB_1802c3872;
  }
  plVar8 = DAT_1803c4b78;
  local_e8 = DAT_1803c4b78;
  FUN_180001d00(param_1,DAT_1803c4b78);
  iVar5 = FUN_180001510(local_98);
  if (iVar5 == -1) {
    plVar9 = *(longlong **)(param_1 + 0x60);
    uVar17 = 0x27;
    *(undefined8 *)(param_1 + 0x60) = 0;
    pcVar15 = pcVar6;
    pcVar16 = pcVar6;
  }
  else if (iVar5 == 0) {
    lVar7 = FUN_1802c1e80(param_1);
/* novels.services.spot_fix_patches#18="No valid patches returned." */
    uVar3 = DAT_1803c4cc0;
    pcVar15 = (code *)0x0;
    pcVar16 = (code *)0x0;
    if (lVar7 == 0) {
/* novels.services.spot_fix_patches#17="SpotFixPatchApplyResult" */
      FUN_180373340(param_1,&local_c8,DAT_1803c4cb8);
      uVar17 = 0x28;
      plVar9 = local_c8;
    }
    else {
      uVar17 = 0x28;
      local_88 = _Py_FalseStruct_exref;
      *(undefined4 *)(plVar8 + 5) = 0x28;
      local_78 = uVar3;
      local_80 = local_a0;
      lVar7 = FUN_18037ccb0(param_1,lVar7,&local_88,DAT_1803c4cc8);
      if (lVar7 != 0) {
        FUN_180001d80(param_1);
LAB_1802c46e6:
        if ((-1 < *(int *)local_a0) &&
           (*(longlong *)local_a0 = *(longlong *)local_a0 + -1, *(longlong *)local_a0 == 0)) {
          (**(code **)(*(longlong *)(local_a0 + 8) + 0x30))();
        }
        if ((int)*local_98 < 0) {
          return lVar7;
        }
        *local_98 = *local_98 + -1;
        if (*local_98 != 0) {
          return lVar7;
        }
        (**(code **)(local_98[1] + 0x30))();
        return lVar7;
      }
      plVar9 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
  }
  else {
    if (*(int *)pcVar11 + 1 != 0) {
      *(int *)pcVar11 = *(int *)pcVar11 + 1;
    }
    local_d0 = DAT_1803ab250;
    if (*(int *)DAT_1803ab250 + 1 != 0) {
      *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
    }
    local_c0 = DAT_1803ab250;
    local_f0 = DAT_1803ab250;
    if (*(int *)DAT_1803ab250 + 1 != 0) {
      *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
      local_f0 = DAT_1803ab250;
    }
    pcVar15 = local_f0;
    DAT_1803ab250 = local_f0;
    if (*(int *)local_f0 + 1 != 0) {
      *(int *)local_f0 = *(int *)local_f0 + 1;
    }
    local_d8 = pcVar11;
    local_90 = (longlong *)FUN_180001930();
    if (local_90 == (longlong *)0x0) {
      plVar9 = *(longlong **)(param_1 + 0x60);
      uVar17 = 0x31;
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar8 = local_e8;
      pcVar16 = pcVar11;
    }
    else {
      uVar17 = 0x32;
      pcVar14 = pcVar6;
      do {
        pcVar10 = local_a8;
        pcVar16 = local_b0;
        pcVar15 = *(code **)(local_90[1] + 0xe0);
        local_e0 = pcVar11;
        if (pcVar15 == (code *)0x0) {
          PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                       *(undefined8 *)(local_90[1] + 0x18));
LAB_1802c4033:
          plVar9 = *(longlong **)(param_1 + 0x60);
          if (plVar9 != (longlong *)0x0) {
            plVar8 = (longlong *)plVar9[1];
            if ((*(uint *)((longlong *)plVar9[1] + 0x15) & 0x40000000) == 0) {
              plVar8 = plVar9;
            }
            if ((plVar8 != *(longlong **)PyExc_StopIteration_exref) &&
               (((-1 < *(int *)(plVar8[1] + 0xa8) || ((*(uint *)(plVar8 + 0x15) & 0x40000000) == 0))
                || (cVar4 = FUN_180382640(), cVar4 != '\x01')))) {
              *(undefined8 *)(param_1 + 0x60) = 0;
              uVar17 = 0x31;
              pcVar15 = local_f0;
              pcVar16 = local_d8;
              if (pcVar14 != (code *)0x0) goto LAB_1802c40c6;
              goto LAB_1802c40df;
            }
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
          }
          if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
             (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
            (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
          }
          if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
            (**(code **)(local_90[1] + 0x30))();
          }
          lVar7 = FUN_1802c1e80(param_1);
          pcVar6 = local_d0;
          pcVar11 = DAT_1803ab250;
          if (lVar7 == 0) {
/* novels.services.spot_fix_patches#17="SpotFixPatchApplyResult" */
            FUN_180373340(param_1,&local_c8,DAT_1803c4cb8);
            uVar17 = 0x39;
            plVar9 = local_c8;
            plVar8 = local_e8;
            pcVar15 = local_f0;
            pcVar16 = local_d8;
            goto LAB_1802c475c;
          }
          if (local_d0 == (code *)0x0) {
            plVar8 = (longlong *)
                     PyUnicode_FromFormat
                               ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#22="applied_count" */
                                ,DAT_1803c4ce0);
            plVar9 = (longlong *)
                     FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8);
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            FUN_18000d1d0(param_1,plVar9);
          }
          else {
            pcVar15 = *(code **)(local_d0 + 8);
            if (pcVar15 == PyLong_Type_exref) {
              pcVar10 = (code *)FUN_180385780(local_d0,DAT_1803ab250);
LAB_1802c42d2:
              if (pcVar10 != (code *)0x0) {
                iVar5 = FUN_180001510(pcVar10);
                pcVar16 = local_d8;
                if (iVar5 == -1) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  goto LAB_1802c4747;
                }
                if (iVar5 == 1) {
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  pcVar16 = local_d8;
                  if (local_d8 == (code *)0x0) {
                    plVar8 = (longlong *)
                             PyUnicode_FromFormat
                                       ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#21="current" */
                                        ,DAT_1803c4cd8);
                    plVar9 = (longlong *)
                             FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,
                                           plVar8);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    FUN_18000d1d0(param_1,plVar9);
                    uVar17 = 0x3a;
                    plVar8 = local_e8;
                    pcVar15 = local_f0;
                    goto LAB_1802c475c;
                  }
                  pcVar10 = (code *)FUN_180385e10(local_d8,local_a0);
                  if (pcVar10 == (code *)0x0) {
                    plVar9 = *(longlong **)(param_1 + 0x60);
                    uVar17 = 0x3a;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar8 = local_e8;
                    pcVar15 = local_f0;
                    goto LAB_1802c475c;
                  }
                }
                else if (local_d8 == (code *)0x0) {
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  plVar8 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#21="current" */
                                      ,DAT_1803c4cd8);
                  plVar9 = (longlong *)
                           FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8
                                        );
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  FUN_18000d1d0(param_1,plVar9);
                  uVar17 = 0x3b;
                  plVar8 = local_e8;
                  pcVar15 = local_f0;
                  goto LAB_1802c475c;
                }
                pcVar11 = local_f0;
                if (local_c0 == (code *)0x0) {
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  plVar8 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#25="skipped_count" */
                                      ,DAT_1803c4cf8);
                  plVar9 = (longlong *)
                           FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8
                                        );
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  FUN_18000d1d0(param_1,plVar9);
                  uVar17 = 0x3d;
                  plVar8 = local_e8;
                  pcVar15 = local_f0;
                  goto LAB_1802c475c;
                }
                if (local_f0 == (code *)0x0) {
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  plVar8 = (longlong *)
                           PyUnicode_FromFormat
                                     ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#23="touched_chars" */
                                      ,DAT_1803c4ce8);
                  plVar9 = (longlong *)
                           FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8
                                        );
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  FUN_18000d1d0(param_1,plVar9);
                  uVar17 = 0x3e;
                  plVar8 = local_e8;
                  pcVar15 = local_f0;
                  goto LAB_1802c475c;
                }
                iVar5 = FUN_1803858f0(pcVar6,DAT_1803ab250);
                if (iVar5 == -1) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < *(int *)pcVar10) &&
                     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                  }
                  uVar17 = 0x3f;
                  plVar8 = local_e8;
                  pcVar15 = local_f0;
                  goto LAB_1802c475c;
                }
/* novels.services.spot_fix_patches#26="No patches could be matched to the chapter content." */
                local_48 = DAT_1803c4d00;
                if (iVar5 == 1) {
                  local_48 = _Py_NoneStruct_exref;
                }
                uVar17 = 0x39;
                *(undefined4 *)(local_e8 + 5) = 0x39;
                local_58 = local_c0;
                local_50 = pcVar11;
                local_70 = pcVar10;
                local_68 = pcVar16;
                local_60 = pcVar6;
                lVar7 = FUN_18037ccb0(param_1,lVar7,&local_70,DAT_1803c4d08);
                if ((-1 < *(int *)pcVar10) &&
                   (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
                  (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                }
                if (lVar7 == 0) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar8 = local_e8;
                  pcVar15 = local_f0;
                  goto LAB_1802c475c;
                }
                FUN_180001d80(param_1);
                pcVar10 = local_b0;
                pcVar14 = local_b8;
                pcVar15 = local_c0;
                if (((pcVar16 != (code *)0x0) && (pcVar6 = local_d0, -1 < *(int *)pcVar16)) &&
                   (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
                  (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
                  pcVar6 = local_d0;
                }
                if ((-1 < *(int *)pcVar6) &&
                   (*(longlong *)pcVar6 = *(longlong *)pcVar6 + -1, *(longlong *)pcVar6 == 0)) {
                  (**(code **)(*(longlong *)(pcVar6 + 8) + 0x30))();
                }
                if ((-1 < *(int *)pcVar15) &&
                   (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
                  (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
                }
                if ((-1 < *(int *)pcVar11) &&
                   (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
                  (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
                }
                if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
                   (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
                  (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
                }
                if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
                   (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
                  (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                }
                goto LAB_1802c46e6;
              }
            }
            else {
              bVar18 = false;
              pcVar14 = PyLong_Type_exref;
              cVar4 = FUN_180382640(PyLong_Type_exref,pcVar15);
              pcVar6 = local_d0;
              if ((cVar4 != '\0') && (*(code **)(pcVar14 + 200) != (code *)0x0)) {
                bVar18 = true;
                pcVar10 = (code *)(**(code **)(pcVar14 + 200))(pcVar11,local_d0,0);
                pcVar14 = PyLong_Type_exref;
                if (pcVar10 != _Py_NotImplementedStruct_exref) goto LAB_1802c42d2;
              }
              if (((*(code **)(pcVar15 + 200) != (code *)0x0) &&
                  (pcVar10 = (code *)(**(code **)(pcVar15 + 200))(pcVar6,pcVar11,4),
                  pcVar14 = PyLong_Type_exref, pcVar6 = local_d0,
                  pcVar10 != _Py_NotImplementedStruct_exref)) ||
                 ((pcVar6 = local_d0, !bVar18 &&
                  ((*(code **)(pcVar14 + 200) != (code *)0x0 &&
                   (pcVar10 = (code *)(**(code **)(pcVar14 + 200))(pcVar11,local_d0,0),
                   pcVar10 != _Py_NotImplementedStruct_exref)))))) goto LAB_1802c42d2;
              PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                           "\'>\' not supported between instances of \'%s\' and \'int\'",
                           *(undefined8 *)(pcVar15 + 0x18));
            }
            plVar9 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
LAB_1802c4747:
          uVar17 = 0x3a;
          plVar8 = local_e8;
          pcVar15 = local_f0;
          pcVar16 = local_d8;
          goto LAB_1802c475c;
        }
        pcVar11 = (code *)(*pcVar15)();
        if (pcVar11 == (code *)0x0) goto LAB_1802c4033;
        local_a8 = pcVar11;
        if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
           (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
          (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
        }
        if (*(int *)pcVar11 + 1 != 0) {
          *(int *)pcVar11 = *(int *)pcVar11 + 1;
        }
        local_b0 = pcVar11;
        if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
           (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
          (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))();
        }
        lVar7 = *(longlong *)(DAT_1803ca470 + 0x20);
        pcVar14 = pcVar11;
        if (*(char *)(lVar7 + 10) == '\0') {
          plVar8 = (longlong *)FUN_1800015a0();
          if (plVar8 == (longlong *)0x0) {
            lVar12 = 0;
          }
          else {
            lVar12 = *plVar8;
          }
LAB_1802c3b4f:
          if (lVar12 == 0) goto LAB_1802c3b54;
        }
        else {
          iVar5 = *(int *)(lVar7 + 0xc);
          if (iVar5 == 0) {
            iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
            *(int *)(lVar7 + 0xc) = iVar5;
          }
          if (iVar5 != DAT_1803a9ed4) {
            DAT_1803a9ed4 = iVar5;
/* novels.services.spot_fix_patches#20="_try_apply_patch" */
            DAT_1803c4c18 = FUN_1803707d0(lVar7,DAT_1803c4cd0,*(undefined8 *)(DAT_1803c4cd0 + 0x18))
            ;
          }
          if (-1 < DAT_1803c4c18) {
            lVar13 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) + lVar7;
            lVar12 = *(longlong *)(lVar13 + 0x28 + DAT_1803c4c18 * 0x10);
            if (lVar12 != 0) goto LAB_1802c3b7c;
/* novels.services.spot_fix_patches#20="_try_apply_patch" */
            DAT_1803c4c18 = FUN_1803707d0(lVar7,DAT_1803c4cd0,*(undefined8 *)(DAT_1803c4cd0 + 0x18))
            ;
            if (-1 < DAT_1803c4c18) {
              lVar12 = *(longlong *)(lVar13 + 0x28 + DAT_1803c4c18 * 0x10);
              goto LAB_1802c3b4f;
            }
          }
LAB_1802c3b54:
/* novels.services.spot_fix_patches#20="_try_apply_patch" */
          plVar8 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c4cd0);
          if ((plVar8 == (longlong *)0x0) || (lVar12 = *plVar8, lVar12 == 0)) {
/* novels.services.spot_fix_patches#20="_try_apply_patch" */
            FUN_180373340(param_1,&local_c8,DAT_1803c4cd0);
            plVar9 = local_c8;
            pcVar15 = local_f0;
            pcVar16 = local_d8;
            goto LAB_1802c40c6;
          }
        }
LAB_1802c3b7c:
        pcVar16 = local_d8;
        if (local_d8 == (code *)0x0) {
          plVar8 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#21="current" */
                              ,DAT_1803c4cd8);
          plVar9 = (longlong *)
                   FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          FUN_18000d1d0(param_1,plVar9);
          pcVar15 = local_f0;
          pcVar16 = local_d8;
          goto LAB_1802c40c6;
        }
        local_88 = local_d8;
        *(undefined4 *)(local_e8 + 5) = 0x32;
        local_80 = pcVar11;
        pcVar11 = (code *)FUN_180377750(param_1,lVar12,&local_88);
        pcVar15 = local_b8;
        if (pcVar11 == (code *)0x0) {
          plVar9 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar15 = local_f0;
          goto LAB_1802c40c6;
        }
        pcVar10 = pcVar11;
        if (((local_b8 != (code *)0x0) && (-1 < *(int *)pcVar6)) &&
           (*(longlong *)pcVar6 = *(longlong *)pcVar6 + -1, *(longlong *)pcVar6 == 0)) {
          pcVar6 = local_b8 + 8;
          local_b8 = pcVar11;
          (**(code **)(*(longlong *)pcVar6 + 0x30))(pcVar15);
          pcVar10 = local_b8;
        }
        local_b8 = pcVar10;
        if (pcVar11 != _Py_NoneStruct_exref) {
          if (*(int *)pcVar11 + 1 != 0) {
            *(int *)pcVar11 = *(int *)pcVar11 + 1;
          }
          local_d8 = pcVar11;
          if ((-1 < *(int *)local_e0) &&
             (*(longlong *)local_e0 = *(longlong *)local_e0 + -1, *(longlong *)local_e0 == 0)) {
            (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
          }
          pcVar16 = pcVar11;
          if (local_d0 == (code *)0x0) {
            plVar8 = (longlong *)
                     PyUnicode_FromFormat
                               ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#22="applied_count" */
                                ,DAT_1803c4ce0);
            plVar9 = (longlong *)
                     FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8);
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            FUN_18000d1d0(param_1,plVar9);
            uVar17 = 0x35;
            pcVar15 = local_f0;
          }
          else {
            local_e0 = local_d0;
            cVar4 = FUN_180382990(&local_e0,DAT_1803ab258);
            if (cVar4 == '\0') {
              plVar9 = *(longlong **)(param_1 + 0x60);
              uVar17 = 0x35;
              *(undefined8 *)(param_1 + 0x60) = 0;
              pcVar15 = local_f0;
            }
            else {
              local_d0 = local_e0;
              if (local_f0 == (code *)0x0) {
                plVar8 = (longlong *)
                         PyUnicode_FromFormat
                                   ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#23="touched_chars" */
                                    ,DAT_1803c4ce8);
                plVar9 = (longlong *)
                         FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8);
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                FUN_18000d1d0(param_1,plVar9);
                uVar17 = 0x36;
                pcVar15 = local_f0;
              }
              else {
                local_e0 = local_f0;
                plVar8 = (longlong *)FUN_18037f0c0();
                if (plVar8 != (longlong *)0x0) {
                  lVar7 = FUN_180373860(plVar8);
                  if ((lVar7 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) {
                    plVar9 = (longlong *)0x0;
                  }
                  else {
                    plVar9 = (longlong *)PyLong_FromSsize_t(lVar7);
                  }
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (plVar9 != (longlong *)0x0) {
                    cVar4 = FUN_180382990(&local_e0,plVar9);
                    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                    }
                    if (cVar4 != '\0') {
                      local_f0 = local_e0;
                      goto LAB_1802c3d43;
                    }
                  }
                }
                plVar9 = *(longlong **)(param_1 + 0x60);
                uVar17 = 0x36;
                *(undefined8 *)(param_1 + 0x60) = 0;
                pcVar15 = local_f0;
              }
            }
          }
          goto LAB_1802c40c6;
        }
        if (local_c0 == (code *)0x0) {
          plVar8 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#25="skipped_count" */
                              ,DAT_1803c4cf8);
          plVar9 = (longlong *)
                   FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          FUN_18000d1d0(param_1,plVar9);
          uVar17 = 0x38;
          pcVar15 = local_f0;
          pcVar16 = local_d8;
          goto LAB_1802c40c6;
        }
        local_e0 = local_c0;
        cVar4 = FUN_180382990(&local_e0,DAT_1803ab258);
        if (cVar4 == '\0') {
          plVar9 = *(longlong **)(param_1 + 0x60);
          uVar17 = 0x38;
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar15 = local_f0;
          pcVar16 = local_d8;
          goto LAB_1802c40c6;
        }
        local_c0 = local_e0;
        pcVar11 = local_d8;
LAB_1802c3d43:
        pcVar15 = local_f0;
        lVar7 = *(longlong *)(param_1 + 0x10);
        if ((((*(int *)(*(longlong *)(lVar7 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar7 + 0x1a0) != 0)
             ) && (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) &&
           (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_1802c3f04;
        if (*(int *)(lVar7 + 0x174) != 0) {
          PyEval_SaveThread();
          PyEval_AcquireThread(param_1);
        }
        pcVar6 = local_b8;
      } while (*(longlong *)(param_1 + 0x80) == 0);
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x80) = 0;
      uVar17 = FUN_180373490(param_1);
      *(undefined8 *)(param_1 + 0x60) = uVar17;
      if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
LAB_1802c3f04:
      plVar9 = *(longlong **)(param_1 + 0x60);
      uVar17 = 0x31;
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar16 = pcVar11;
LAB_1802c40c6:
      if ((-1 < *(int *)pcVar14) &&
         (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
        (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
      }
LAB_1802c40df:
      plVar8 = local_e8;
      if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
        (**(code **)(local_90[1] + 0x30))();
        plVar8 = local_e8;
      }
    }
  }
LAB_1802c475c:
  piVar1 = (int *)plVar9[5];
  if (piVar1 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar8,uVar17);
    bVar18 = pcVar11 == _Py_NoneStruct_exref;
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar8) goto LAB_1802c47fe;
    pcVar11 = (code *)FUN_18038e2a0(plVar8,uVar17);
    *(int **)(pcVar11 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    bVar18 = pcVar11 == _Py_NoneStruct_exref;
  }
  plVar8 = (longlong *)plVar9[5];
  if (bVar18) {
    pcVar6 = (code *)0x0;
  }
  else {
    pcVar6 = pcVar11;
    if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
      *(int *)pcVar11 = *(int *)pcVar11 + 1;
    }
  }
  plVar9[5] = (longlong)pcVar6;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  plVar8 = local_e8;
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
    plVar8 = local_e8;
  }
LAB_1802c47fe:
  pcVar6 = local_c0;
  pcVar11 = local_d0;
  local_100 = local_b8;
  local_108 = local_b0;
  local_118 = local_c0;
  local_120 = local_d0;
  local_128 = pcVar16;
  local_110 = pcVar15;
  FUN_18036d0a0(plVar8,"oooooooo",local_a0,local_98);
  plVar2 = DAT_1803c4b78;
  if (plVar8 == DAT_1803c4b78) {
    if ((-1 < (int)*DAT_1803c4b78) && (*DAT_1803c4b78 = *DAT_1803c4b78 + -1, *plVar2 == 0)) {
      (**(code **)(plVar2[1] + 0x30))();
    }
    DAT_1803c4b78 = (longlong *)0x0;
  }
  lVar7 = **(longlong **)(param_1 + 0x38);
  plVar8 = *(longlong **)(lVar7 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar7 + 8);
  *(undefined4 *)(plVar8 + 8) = 0xffffffff;
  plVar2 = (longlong *)plVar8[2];
  if (((plVar2 != (longlong *)0x0) && (plVar8[2] = 0, -1 < (int)*plVar2)) &&
     (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  plVar8[10] = 0;
  if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
     (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
  if (((pcVar6 != (code *)0x0) && (-1 < *(int *)pcVar6)) &&
     (*(longlong *)pcVar6 = *(longlong *)pcVar6 + -1, *(longlong *)pcVar6 == 0)) {
    (**(code **)(*(longlong *)(pcVar6 + 8) + 0x30))(pcVar6);
  }
  if (((pcVar15 != (code *)0x0) && (-1 < *(int *)pcVar15)) &&
     (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
    (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
  }
  if (((local_b0 != (code *)0x0) && (-1 < *(int *)local_b0)) &&
     (*(longlong *)local_b0 = *(longlong *)local_b0 + -1, *(longlong *)local_b0 == 0)) {
    (**(code **)(*(longlong *)(local_b0 + 8) + 0x30))();
  }
  if (((local_b8 != (code *)0x0) && (-1 < *(int *)local_b8)) &&
     (*(longlong *)local_b8 = *(longlong *)local_b8 + -1, *(longlong *)local_b8 == 0)) {
    (**(code **)(*(longlong *)(local_b8 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_a0) &&
     (*(longlong *)local_a0 = *(longlong *)local_a0 + -1, *(longlong *)local_a0 == 0)) {
    (**(code **)(*(longlong *)(local_a0 + 8) + 0x30))();
  }
  if ((-1 < (int)*local_98) && (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  plVar8 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar9;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  return 0;
}
